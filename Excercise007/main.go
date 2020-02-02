package main

import (
	"fmt"
	"unsafe"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

func main() {
	glfw.Init()
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	vk.Init()

	var appInfo *vk.ApplicationInfo = &vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.MakeVersion(1, 0, 0), // Throws 'vulkan error: incompatible driver' error with incorrect version number
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	// Resources
	var instance vk.Instance
	var physicalDevices []vk.PhysicalDevice
	var physicalDeviceProperties vk.PhysicalDeviceProperties
	var physicalDeviceFeatures vk.PhysicalDeviceFeatures
	var memoryProperties vk.PhysicalDeviceMemoryProperties
	var pQueueFamilyProperties []vk.QueueFamilyProperties
	var pLogicalDevice *vk.Device
	var commandPool *vk.CommandPool
	var commandBuffers []vk.CommandBuffer
	var pBuffer, dstBuffer *vk.Buffer
	var imageFormatProperties vk.ImageFormatProperties
	var pImageBuffer *vk.Image
	var pHostMemory unsafe.Pointer
	var pDeviceMemory *vk.DeviceMemory
	var pImageView *vk.ImageView
	var pQueue *vk.Queue

	//Create Instance
	var layers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	// For Windows Only
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	// For Linux
	//https://software.intel.com/en-us/articles/api-without-secrets-introduction-to-vulkan-part-2
	//var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_xcb_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PNext: nil,
		//Flags:                   nil,
		PApplicationInfo:        appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	result := vk.CreateInstance(&instanceInfo, nil, &instance)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating instance: %v", result))
	}

	var physicalDeviceIndex int = 1 // 0= NVIDIA Geforce MX150, 1=Intel(R) UHD Graphics 620 [On my HP Laptop]

	physicalDevices = getPhysicalDevices(instance)
	physicalDeviceProperties = getPhysicalDeviceProperties(physicalDevices[physicalDeviceIndex])
	physicalDeviceFeatures = getPhysicalDeviceFeatures(physicalDevices[physicalDeviceIndex])
	memoryProperties = getPhysicalDeviceMemoryProperties(physicalDevices[physicalDeviceIndex])
	printPhysicalDeviceMemoryProperties(memoryProperties)
	pQueueFamilyProperties = getPhysicalDeviceQueueFamilyProperties(physicalDevices[physicalDeviceIndex])
	pLogicalDevice = createDevice(physicalDevices[physicalDeviceIndex], physicalDeviceFeatures)
	getInstanceLayerProperties()
	getDeviceLayerProperties(physicalDevices[physicalDeviceIndex])
	getInstanceExtensionProperties()
	getDeviceExtensionProperties(physicalDevices[physicalDeviceIndex])
	deviceWaitTillComplete(pLogicalDevice)

	pBuffer = createBuffer(*pLogicalDevice)
	imageFormatProperties = getPhysicalDeviceImageProperties(physicalDevices[physicalDeviceIndex])
	pImageBuffer = createImageBuffer(pLogicalDevice)
	// List Supported Image Format by GPU
	//checkSupportedImageFormat(physicalDevices[physicalDeviceIndex])
	getBufferMemoryRequirements(*pLogicalDevice, *pBuffer)
	pHostMemory, pDeviceMemory = mapHostMemoryForImage(*pLogicalDevice, memoryProperties, *pImageBuffer)
	bindImageMemory(*pLogicalDevice, *pImageBuffer, pDeviceMemory)
	pImageView = createImageView(*pLogicalDevice, *pImageBuffer)
	pQueue = getDeviceQueue(*pLogicalDevice)
	//pBufferView = createBufferView(*pLogicalDevice, *pBuffer)
	// Get the memory properties of the physical device.
	vk.GetPhysicalDeviceMemoryProperties(physicalDevices[physicalDeviceIndex], &memoryProperties)

	//Command Buffer recording
	commandPool = createCommandPool(*pLogicalDevice)
	commandBuffers = allocateCommandBuffers(*pLogicalDevice, *commandPool, 2)
	beginCommandBuffer(commandBuffers)

	// Verbose - Please don't remove, ignore
	physicalDeviceProperties.Deref()
	fmt.Println("\n===========================\n    OUTPUTS    \n===========================")
	fmt.Println("Physical Devices present....", physicalDevices)
	//Physical Device name
	fmt.Println(vk.ToString(physicalDeviceProperties.DeviceName[:]))
	fmt.Println(physicalDeviceFeatures)
	fmt.Println(memoryProperties)
	fmt.Println(pQueueFamilyProperties)
	printDeviceQueueFamilyProperties(pQueueFamilyProperties)
	fmt.Printf("%T, %v", pLogicalDevice, pLogicalDevice)
	fmt.Println(pBuffer, dstBuffer)
	fmt.Println(&imageFormatProperties)
	fmt.Println(commandPool)
	fmt.Println(commandBuffers)
	fmt.Println(pImageBuffer)
	fmt.Println("Host Memory Pointer ", pHostMemory)
	fmt.Println("Image Buffer View Pointer ", pImageView)
	fmt.Println("Device Queue......", pQueue)

	//Cleaningup code
	vk.FreeCommandBuffers(*pLogicalDevice, *commandPool, 1, commandBuffers)
	vk.DestroyCommandPool(*pLogicalDevice, *commandPool, nil)
}

func recordCommandIntoCommandBuffer(commandBuffer vk.CommandBuffer) {}
func beginCommandBuffer(commandBuffer []vk.CommandBuffer) {
	fmt.Println("Begin Command Buffers.................")
	var commandBufferBeginInfo = vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: 0x0,
	}
	for _, cmdBuffer := range commandBuffer {
		result := vk.BeginCommandBuffer(cmdBuffer, &commandBufferBeginInfo)
		if result != vk.Success {
			fmt.Printf("Failed to begin command buffer with error : %v", result)
		}
	}
}

func allocateCommandBuffers(pLogicalDevice vk.Device, pCommandPool vk.CommandPool, count uint32) []vk.CommandBuffer {
	fmt.Println("Allocating command buffer...............")
	var commandBuffers []vk.CommandBuffer
	commandBuffers = make([]vk.CommandBuffer, count)
	var commandBufferAllocateInfo = vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        pCommandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: count,
	}
	result := vk.AllocateCommandBuffers(pLogicalDevice, &commandBufferAllocateInfo, commandBuffers)
	if result != vk.Success {
		fmt.Printf("Failed to create command buffer with error : %v", result)
	}
	return commandBuffers
}

func createCommandPool(pLogicalDevice vk.Device) *vk.CommandPool {
	fmt.Println("Creating Command Pool........")
	// Create CommandPool first
	var commandPool vk.CommandPool
	var commandPollCreateInfo = vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit | vk.CommandPoolCreateTransientBit),
		QueueFamilyIndex: 0,
	}
	result := vk.CreateCommandPool(pLogicalDevice, &commandPollCreateInfo, nil, &commandPool)
	if result != vk.Success {
		fmt.Printf("Failed to create command pool with error : %v", result)
	}
	return &commandPool
}

func getDeviceQueue(pLogicalDevice vk.Device) *vk.Queue {
	fmt.Println("Getting Device Queue................")
	var queue vk.Queue
	vk.GetDeviceQueue(pLogicalDevice, 0, 0, &queue)
	return &queue
}

func createImageView(pLogicalDevice vk.Device, imageBuffer vk.Image) *vk.ImageView {
	fmt.Println("Creating Image View........")
	var imageView vk.ImageView
	var imageViewCreateInfo = vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    imageBuffer,
		ViewType: vk.ImageViewType2d,     // It must be compatible with Image Buffer's ImageType in ImageCreateInfo struct
		Format:   vk.FormatR8g8b8a8Unorm, // It's same as Format in Image Buffer's ImageCreateInfo struct
		Components: vk.ComponentMapping{
			R: vk.ComponentSwizzleR,
			G: vk.ComponentSwizzleG,
			B: vk.ComponentSwizzleB,
			A: vk.ComponentSwizzleA,
		},
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			LevelCount: 1,
			LayerCount: 1,
		},
	}
	result := vk.CreateImageView(pLogicalDevice, &imageViewCreateInfo, nil, &imageView)
	if result != vk.Success {
		fmt.Printf("Failed to map memory to image buffer with error : %v", result)
		return nil
	}
	return &imageView
}

func bindImageMemory(pLogicalDevice vk.Device, imageBuffer vk.Image, pHostMemory *vk.DeviceMemory) {
	// Before a resource such as a buffer or image can be used by Vulkan to store data, memory must be
	// bound to it. Before memory is bound to a resource, you should determine what type of memory and
	// how much of it the resource requires. There is a different function for buffers and for textures.
	// They are vkGetBufferMemoryRequirements() and vkGetImageMemoryRequirements()
	// The only difference between these two functions is that vkGetBufferMemoryRequirements() takes a
	// handle to a buffer object and vkGetImageMemoryRequirements() takes a handle to an image object.
	fmt.Println("Binding Device Memory............")
	result := vk.BindImageMemory(pLogicalDevice, imageBuffer, *pHostMemory, vk.DeviceSize(0))
	if result != vk.Success {
		fmt.Printf("Failed to failed to bind memory to image with error : %v", result)
	}
}

// func createBufferView(pLogicalDevice vk.Device, buffer vk.Buffer) *vk.BufferView {
// 	var bufferView vk.BufferView
// 	var bufferViewCreateInfo = vk.BufferViewCreateInfo{
// 		SType:  vk.StructureTypeBufferViewCreateInfo,
// 		PNext:  nil,
// 		Buffer: buffer,
// 		Offset: 0,
// 		Range:  1024, //Just uing 1Kb only
// 	}
// 	result := vk.CreateBufferView(pLogicalDevice, &bufferViewCreateInfo, nil, &bufferView)
// 	if result != vk.Success {
// 		fmt.Printf("Failed to create buffer view with error : %v", result)
// 		return nil
// 	}
// 	return &bufferView
// }

// func bindMemoryToBuffer() {
// 	vertexData := linmath.ArrayFloat32([]float32{
// 		-1, -1, 0,
// 		1, -1, 0,
// 		0, 1, 0,
// 	})

// }

func mapHostMemoryForImage(pLogicalDevice vk.Device, memoryProperties vk.PhysicalDeviceMemoryProperties, imageBuffer vk.Image) (unsafe.Pointer, *vk.DeviceMemory) {
	// Access to this memory object must be externally synchronized
	// A flush is necessary if the host has written to a mapped memory
	// region and needs the device to see the effect of those writes.
	// However, if the device writes to a mapped memory region and you
	// need the host to see the effect of the device’s writes, you need
	// to invalidate any caches on the host that might now hold stale data.
	// To do this, call vkInvalidateMappedMemoryRanges()
	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(pLogicalDevice, imageBuffer, &memReqs)
	memReqs.Deref()
	memoryProperties.Deref()

	var pData unsafe.Pointer
	memAlloc := &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: 0, //MemoryTypeIndex is an index into the memory type array returned from a call to vkGetPhysicalDeviceMemoryProperties()
		// I found index of memory contaning 'MemoryPropertyDeviceLocalBit' and 'MemoryPropertyHostVisibleBit', from the output of printPhysicalDeviceMemoryProperties() function
	}
	var mem vk.DeviceMemory
	result := vk.AllocateMemory(pLogicalDevice, memAlloc, nil, &mem)
	if result != vk.Success {
		fmt.Printf("Failed to map memory to image buffer with error : %v", result)
		return nil, nil
	}
	vk.MapMemory(pLogicalDevice, mem, vk.DeviceSize(0), vk.DeviceSize(vk.WholeSize), 0, &pData)
	return pData, &mem
}

// func mapHostMemoryForBuffer() {

// }

func createImageBuffer(pLogicalDevice *vk.Device) *vk.Image {
	fmt.Println("Creating image buffer...........")
	var imageBuffer vk.Image
	var extent = vk.Extent3D{
		Width:  1024,
		Height: 1024,
		Depth:  1,
	}
	//=================
	// CUBE MAPS (cube map & cube-map array image)
	//=================
	// Set following(top 3) in ImageCreateInfo struct and below 4,5,6 in ImageViewCreateInfo struct and below 7,8,9 inside SubresourceRange section of ImageViewCreateInfo to create a Cube Map:
	// ImageCreateInfo------------------------
	// ArrayLayers: 6
	// TimageType to vk.ImageType2d
	// Flags: vk.ImageCreateCubeCompatibleBit OR
	// ImageViewCreateInfo --------------------
	// ViewType: vk.ImageViewTypeCubeArray
	// ViewType: vk.ImageViewTypeCube // we create a view of the 2D array parent, but rather than creating a normal 2D (array) view of the image, we create a cube-map view
	// layerCount : 6 #Cube maps can also form arrays of their own. This is simply a concatenation of an integer multiple of six faces, with each group of six forming a separate cube. To create a cube-map array image, set the viewType field of VkImageViewCreateInfo to VK_IMAGE_VIEW_TYPE_CUBE_ARRAY
	// baseArrayLayer: #numberOfCubesToMake
	// layerCount: 6, //To create a single cube, layerCount should be set to 6

	var imageCreateInfo = vk.ImageCreateInfo{
		SType:                 vk.StructureTypeImageCreateInfo,
		ImageType:             vk.ImageType2d,
		Format:                vk.FormatR8g8b8a8Unorm, //vk.FormatR8g8b8a8Srgb,
		Extent:                extent,
		MipLevels:             10,
		ArrayLayers:           1,
		Samples:               vk.SampleCount1Bit,
		Tiling:                vk.ImageTilingOptimal,
		Usage:                 vk.ImageUsageFlags(vk.ImageUsageSampledBit),
		SharingMode:           vk.SharingModeExclusive,
		QueueFamilyIndexCount: 0,
		InitialLayout:         vk.ImageLayoutUndefined,
	}
	result := vk.CreateImage(*pLogicalDevice, &imageCreateInfo, nil, &imageBuffer)
	if result != vk.Success {
		fmt.Printf("Failed to create image buffer with error : %v", result)
	}
	return &imageBuffer
}

func getBufferMemoryRequirements(pLogicalDevice vk.Device, buffer vk.Buffer) {
	var memoryProperties = map[string]vk.MemoryPropertyFlagBits{
		"MemoryPropertyDeviceLocalBit":     vk.MemoryPropertyDeviceLocalBit,
		"MemoryPropertyHostVisibleBit":     vk.MemoryPropertyHostVisibleBit,
		"MemoryPropertyHostCoherentBit":    vk.MemoryPropertyHostCoherentBit,
		"MemoryPropertyHostCachedBit":      vk.MemoryPropertyHostCachedBit,
		"MemoryPropertyLazilyAllocatedBit": vk.MemoryPropertyLazilyAllocatedBit,
		"MemoryPropertyProtectedBit":       vk.MemoryPropertyProtectedBit,
		"MemoryPropertyFlagBitsMaxEnum":    vk.MemoryPropertyFlagBitsMaxEnum,
	}
	fmt.Println("Printing Buffer memory requirements..........")
	var memoryRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(pLogicalDevice, buffer, &memoryRequirements)
	memoryRequirements.Deref()
	fmt.Println("\t\t* Size Required ...", memoryRequirements.Size, " Bytes")
	fmt.Println("\t\t* Alignment Required ...", memoryRequirements.Alignment, " Bytes")
	fmt.Println("\t\t* MemoryTypes Required ...")
	for key, mp := range memoryProperties {
		if mp&vk.MemoryPropertyFlagBits(memoryRequirements.MemoryTypeBits) != 0x00000000 {
			fmt.Println("\t\t\t\t* ", key)
		}
	}
}

func getPhysicalDeviceImageProperties(physicalDevice vk.PhysicalDevice) vk.ImageFormatProperties {
	fmt.Println("Physical device supported image format for 3D data of sRGB FormatB8g8r8Unorm Format........") // Another popular format vk.FormatR8g8b8Unorm
	var imageType vk.ImageType = vk.ImageType3d
	var imageTiling vk.ImageTiling = vk.ImageTilingLinear
	//var imageUsageFlags vk.ImageUsageFlagBits
	var imageCreateFlags vk.ImageCreateFlags
	var imageFormatProperties vk.ImageFormatProperties
	result := vk.GetPhysicalDeviceImageFormatProperties(physicalDevice, vk.FormatR8g8b8a8Unorm, imageType, imageTiling, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit), imageCreateFlags, &imageFormatProperties)
	if result != vk.Success {
		fmt.Printf("Failed to get Physical device supported image formats with error : %v", result)
	}
	fmt.Printf("\t*\tImageType: %v\t\t \t\t\n", imageType)
	fmt.Printf("\t*\tImageTiling: %v\t\t \t\t\n", imageTiling)
	imageFormatProperties.Deref()
	extent := imageFormatProperties.MaxExtent
	maxmipmaplevel := imageFormatProperties.MaxMipLevels
	maxarraylayers := imageFormatProperties.MaxArrayLayers
	maxResourceSize := imageFormatProperties.MaxResourceSize
	extent.Deref()
	fmt.Printf("\t*\tExtent: \t\t \t\t\n")
	fmt.Printf("\t\t\tWidth: %v\t\t\n", extent.Width)
	fmt.Printf("\t\t\tHeight: %v\t\t\n", extent.Height)
	fmt.Printf("\t\t\tDepth: %v\t\t\n", extent.Depth)
	fmt.Printf("\t*\tMaxMipMap Levels: %v\t\t \t\t\n", maxmipmaplevel)
	fmt.Printf("\t*\tMaxArrayLayers: %v\t\t \t\t\n", maxarraylayers)
	fmt.Printf("\t*\tMaxResourceSize: %v\t\t \t\t\n", maxResourceSize)
	return imageFormatProperties
}

func createBuffer(device vk.Device) *vk.Buffer {
	fmt.Println("Creating buffer..........")
	var buffer vk.Buffer
	var bufferCreateInfo = vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Flags:       0x0,
		Size:        1024 * 1024, // 1Mb or vk.DeviceSize(vertexData.Sizeof()) in cause 'Usage' is set vk.BufferUsageVertexBufferBit
		Usage:       vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit | vk.BufferUsageTransferDstBit),
		SharingMode: vk.SharingModeExclusive,
	}
	result := vk.CreateBuffer(device, &bufferCreateInfo, nil, &buffer)
	if result != vk.Success {
		fmt.Printf("Failed to create buffer : %v", result)
	}
	return &buffer
}

func deviceWaitTillComplete(pLogicalDevice *vk.Device) {
	// Wait till the device is finished executing any work on behalf of your application
	// Wait on the host for the completion of outstanding queue operations for all queues on a given logical device
	fmt.Println("Waiting for the device to complete all pending queue operations.....")
	result := vk.DeviceWaitIdle(*pLogicalDevice)
	if result != vk.Success {
		fmt.Printf("Failed to wait for device to complete the execution with error : %v", result)
	}
	fmt.Println("All the queue operations are complete, device is free now....")
}
func getDeviceExtensionProperties(physicalDevice vk.PhysicalDevice) {
	fmt.Println("Listing available Extensions for device only..............")
	var propertyCount uint32
	vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &propertyCount, nil)
	var pProperties []vk.ExtensionProperties = make([]vk.ExtensionProperties, propertyCount)
	vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tExtensionName: %v\n\t\tSpecVersion: %v\n", vk.ToString(p.ExtensionName[:]), p.SpecVersion)
	}
}

func getInstanceExtensionProperties() {
	fmt.Println("Listing available Extensions for instance only..............")
	var propertyCount uint32
	//vk.EnumerateInstanceLayerProperties(&propertyCount, nil)
	vk.EnumerateInstanceExtensionProperties("", &propertyCount, nil)
	var pProperties []vk.ExtensionProperties = make([]vk.ExtensionProperties, propertyCount)
	//vk.EnumerateInstanceLayerProperties(&propertyCount, pProperties)
	vk.EnumerateInstanceExtensionProperties("", &propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tExtensionName: %v\n\t\tSpecVersion: %v\n", vk.ToString(p.ExtensionName[:]), p.SpecVersion)
	}
}

func getDeviceLayerProperties(physicalDevice vk.PhysicalDevice) {
	fmt.Println("Listing available layers for device only..............")
	var propertyCount uint32
	vk.EnumerateDeviceLayerProperties(physicalDevice, &propertyCount, nil)
	var pProperties []vk.LayerProperties = make([]vk.LayerProperties, propertyCount)
	vk.EnumerateDeviceLayerProperties(physicalDevice, &propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tName: %v\n\t\tDescription: %v\n\t\tSpecVersion: %v\n\t\tImplementationVersion: %v\n", vk.ToString(p.LayerName[:]), vk.ToString(p.Description[:]), p.SpecVersion, p.ImplementationVersion)
	}
}

func getInstanceLayerProperties() {
	fmt.Println("Listing available layers for instance only..............")
	var propertyCount uint32
	vk.EnumerateInstanceLayerProperties(&propertyCount, nil)
	var pProperties []vk.LayerProperties = make([]vk.LayerProperties, propertyCount)
	vk.EnumerateInstanceLayerProperties(&propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tName: %v\n\t\tDescription: %v\n\t\tSpecVersion: %v\n\t\tImplementationVersion: %v\n", vk.ToString(p.LayerName[:]), vk.ToString(p.Description[:]), p.SpecVersion, p.ImplementationVersion)
	}
}

func createDevice(physicalDevice vk.PhysicalDevice, physicalDeviceFeatures vk.PhysicalDeviceFeatures) *vk.Device {
	fmt.Println("Creating Logical device..............")
	var logicalDevice vk.Device
	// var physicalDeviceFeatures []vk.PhysicalDeviceFeatures = []vk.PhysicalDeviceFeatures{
	// 	physicalDeviceFeatures,
	// }
	var pEnabledDeviceFeatures []vk.PhysicalDeviceFeatures = make([]vk.PhysicalDeviceFeatures, 1)
	deviceQueueCreateInfoSlice := []vk.DeviceQueueCreateInfo{{
		SType: vk.StructureTypeDeviceQueueCreateInfo,
		//QueueCount:       16,
		QueueCount:       1,
		QueueFamilyIndex: 0,
		PQueuePriorities: []float32{1.0},
	}}
	//var deviceExtensions = []string{"VK_KHR_surface\x00"}
	//var deviceExtensions = []string{"VK_KHR_swapchain\x00"}
	//var deviceLayers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var deviceCreateInfo *vk.DeviceCreateInfo = &vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount: uint32(len(deviceQueueCreateInfoSlice)),
		PQueueCreateInfos:    deviceQueueCreateInfoSlice,
		// EnabledLayerCount:       uint32(len(deviceLayers)),
		// PpEnabledLayerNames:     deviceLayers,
		// EnabledExtensionCount:   uint32(len(deviceExtensions)),
		// PpEnabledExtensionNames: deviceExtensions,
		// Following will enable given features from slice
		//PEnabledFeatures:        physicalDeviceFeatures,
		// Following will return structure of enable features into the slice
		PEnabledFeatures: pEnabledDeviceFeatures,
	}
	result := vk.CreateDevice(physicalDevice, deviceCreateInfo, nil, &logicalDevice)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating Logical device: %v", result))
		return nil
	}
	fmt.Println("Created Logical device..............")
	return &logicalDevice
}

func printDeviceQueueFamilyProperties(qfs []vk.QueueFamilyProperties) {
	fmt.Println("Print device queue properties..............")
	fmt.Printf("\tFound %v families\n", len(qfs))
	for idx, qf := range qfs {
		qf.Deref()
		fmt.Printf("\tFamily %v, Number of queues in the family (vk.QueueCount): %v\n", idx, qf.QueueCount)
		flagBits := vk.QueueFlagBits(qf.QueueFlags)
		if flagBits&vk.QueueGraphicsBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_GRAPHICS_BIT \t\t[vk.QueueGraphicsBit] \t\t[0x00000001\\1]\n")
		}
		if flagBits&vk.QueueComputeBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_COMPUTE_BIT \t\t[vk.QueueComputeBit] \t\t[0x00000002\\2]\n")
		}
		if flagBits&vk.QueueTransferBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_TRANSFER_BIT \t\t[vk.QueueTransferBit] \t\t[0x00000004\\4]\n")
		}
		if flagBits&vk.QueueSparseBindingBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_SPARSE_BINDING_BIT \t[vk.QueueSparseBindingBit] \t[0x00000008\\8]\n")
		}
		if flagBits&vk.QueueProtectedBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_PROTECTED_BIT \t\t[vk.QueueProtectedBit] \t\t[0x00000010\\16]\n")
		}
		if flagBits&vk.QueueFlagBitsMaxEnum != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_FLAG_BITS_MAX_ENUM \t[QueueFlagBitsMaxEnumt] \t[0x7FFFFFFF\\2147483647]\n")
		}
	}
}

func getPhysicalDeviceQueueFamilyProperties(physicalDevice vk.PhysicalDevice) []vk.QueueFamilyProperties {
	var pQueueFamilyPropertyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(physicalDevice, &pQueueFamilyPropertyCount, nil)
	var pQueueFamilyProperties []vk.QueueFamilyProperties
	pQueueFamilyProperties = make([]vk.QueueFamilyProperties, pQueueFamilyPropertyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(physicalDevice, &pQueueFamilyPropertyCount, pQueueFamilyProperties)
	return pQueueFamilyProperties
}

func printPhysicalDeviceMemoryProperties(memoryProperties vk.PhysicalDeviceMemoryProperties) {
	fmt.Println("Printing Device memory properties.......................")
	memoryProperties.Deref()
	fmt.Printf("\t*\tMemoryTypeCount: %v\t\t \t\t\n", memoryProperties.MemoryTypeCount)
	fmt.Printf("\t*\tMemoryHeapCount: %v\t\t \t\t\n", memoryProperties.MemoryHeapCount)
	var memPropertiesFlags = map[string]vk.MemoryPropertyFlagBits{
		"MemoryPropertyDeviceLocalBit":     vk.MemoryPropertyDeviceLocalBit,
		"MemoryPropertyHostVisibleBit":     vk.MemoryPropertyHostVisibleBit,
		"MemoryPropertyHostCoherentBit":    vk.MemoryPropertyHostCoherentBit,
		"MemoryPropertyHostCachedBit":      vk.MemoryPropertyHostCachedBit,
		"MemoryPropertyLazilyAllocatedBit": vk.MemoryPropertyLazilyAllocatedBit,
		"MemoryPropertyProtectedBit":       vk.MemoryPropertyProtectedBit,
		"MemoryPropertyFlagBitsMaxEnum":    vk.MemoryPropertyFlagBitsMaxEnum,
	}
	fmt.Println("Total MemoryTypes Found :", len(memoryProperties.MemoryTypes))
	for idx, memoryType := range memoryProperties.MemoryTypes { //HeapIndex
		memoryType.Deref()
		fmt.Println("\t* MemoryTypes index: ", idx)
		fmt.Printf("\t\t*\tMemoryType HeapIndex: %v\t\t \t\t\n", memoryType.HeapIndex)
		fmt.Printf("\t\t*\tMemoryType PropertyFlags: \t\t \t\t\n")
		for key, memoryProperty := range memPropertiesFlags {
			if memoryProperty&vk.MemoryPropertyFlagBits(memoryType.PropertyFlags) != 0x00000000 {
				fmt.Printf("\t\t\t* Memory Property Flags: %v\t\t \t\t\n", key)
			}
		}
	}
}

func getPhysicalDeviceMemoryProperties(physicalDevice vk.PhysicalDevice) vk.PhysicalDeviceMemoryProperties {
	var pMemoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(physicalDevice, &pMemoryProperties)
	return pMemoryProperties
}

func getPhysicalDeviceFeatures(physicalDevice vk.PhysicalDevice) vk.PhysicalDeviceFeatures {
	var pFeatures vk.PhysicalDeviceFeatures
	vk.GetPhysicalDeviceFeatures(physicalDevice, &pFeatures)
	return pFeatures
}

func getPhysicalDeviceProperties(physicalDevice vk.PhysicalDevice) vk.PhysicalDeviceProperties {
	var pProperties vk.PhysicalDeviceProperties
	vk.GetPhysicalDeviceProperties(physicalDevice, &pProperties)
	return pProperties
}

func getPhysicalDevices(instance vk.Instance) []vk.PhysicalDevice {
	fmt.Println("[PhysicalDevice] Getting information....")
	var deviceCount uint32
	result := vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		return nil
	}
	fmt.Printf("\tFound total %v Physical Device(s).....\n", deviceCount)
	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		return nil
	}
	return physicalDevices
}

func checkSupportedImageFormat(physicalDevice vk.PhysicalDevice) {
	fmt.Println("Listing all the supported image formats.............................")
	var formats = map[string]vk.Format{
		"FormatR4g4UnormPack8":                       vk.FormatR4g4UnormPack8,
		"FormatR4g4b4a4UnormPack16":                  vk.FormatR4g4b4a4UnormPack16,
		"FormatB4g4r4a4UnormPack16":                  vk.FormatB4g4r4a4UnormPack16,
		"FormatR5g6b5UnormPack16":                    vk.FormatR5g6b5UnormPack16,
		"FormatB5g6r5UnormPack16":                    vk.FormatB5g6r5UnormPack16,
		"FormatR5g5b5a1UnormPack16":                  vk.FormatR5g5b5a1UnormPack16,
		"FormatB5g5r5a1UnormPack16":                  vk.FormatB5g5r5a1UnormPack16,
		"FormatA1r5g5b5UnormPack16":                  vk.FormatA1r5g5b5UnormPack16,
		"FormatR8Unorm":                              vk.FormatR8Unorm,
		"FormatR8Snorm":                              vk.FormatR8Snorm,
		"FormatR8Uscaled":                            vk.FormatR8Uscaled,
		"FormatR8Sscaled":                            vk.FormatR8Sscaled,
		"FormatR8Uint":                               vk.FormatR8Uint,
		"FormatR8Sint":                               vk.FormatR8Sint,
		"FormatR8Srgb":                               vk.FormatR8Srgb,
		"FormatR8g8Unorm":                            vk.FormatR8g8Unorm,
		"FormatR8g8Snorm":                            vk.FormatR8g8Snorm,
		"FormatR8g8Uscaled":                          vk.FormatR8g8Uscaled,
		"FormatR8g8Sscaled":                          vk.FormatR8g8Sscaled,
		"FormatR8g8Uint":                             vk.FormatR8g8Uint,
		"FormatR8g8Sint":                             vk.FormatR8g8Sint,
		"FormatR8g8Srgb":                             vk.FormatR8g8Srgb,
		"FormatR8g8b8Unorm":                          vk.FormatR8g8b8Unorm,
		"FormatR8g8b8Snorm":                          vk.FormatR8g8b8Snorm,
		"FormatR8g8b8Uscaled":                        vk.FormatR8g8b8Uscaled,
		"FormatR8g8b8Sscaled":                        vk.FormatR8g8b8Sscaled,
		"FormatR8g8b8Uint":                           vk.FormatR8g8b8Uint,
		"FormatR8g8b8Sint":                           vk.FormatR8g8b8Sint,
		"FormatR8g8b8Srgb":                           vk.FormatR8g8b8Srgb,
		"FormatB8g8r8Unorm":                          vk.FormatB8g8r8Unorm,
		"FormatB8g8r8Snorm":                          vk.FormatB8g8r8Snorm,
		"FormatB8g8r8Uscaled":                        vk.FormatB8g8r8Uscaled,
		"FormatB8g8r8Sscaled":                        vk.FormatB8g8r8Sscaled,
		"FormatB8g8r8Uint":                           vk.FormatB8g8r8Uint,
		"FormatB8g8r8Sint":                           vk.FormatB8g8r8Sint,
		"FormatB8g8r8Srgb":                           vk.FormatB8g8r8Srgb,
		"FormatR8g8b8a8Unorm":                        vk.FormatR8g8b8a8Unorm,
		"FormatR8g8b8a8Snorm":                        vk.FormatR8g8b8a8Snorm,
		"FormatR8g8b8a8Uscaled":                      vk.FormatR8g8b8a8Uscaled,
		"FormatR8g8b8a8Sscaled":                      vk.FormatR8g8b8a8Sscaled,
		"FormatR8g8b8a8Uint":                         vk.FormatR8g8b8a8Uint,
		"FormatR8g8b8a8Sint":                         vk.FormatR8g8b8a8Sint,
		"FormatR8g8b8a8Srgb":                         vk.FormatR8g8b8a8Srgb,
		"FormatB8g8r8a8Unorm":                        vk.FormatB8g8r8a8Unorm,
		"FormatB8g8r8a8Snorm":                        vk.FormatB8g8r8a8Snorm,
		"FormatB8g8r8a8Uscaled":                      vk.FormatB8g8r8a8Uscaled,
		"FormatB8g8r8a8Sscaled":                      vk.FormatB8g8r8a8Sscaled,
		"FormatB8g8r8a8Uint":                         vk.FormatB8g8r8a8Uint,
		"FormatB8g8r8a8Sint":                         vk.FormatB8g8r8a8Sint,
		"FormatB8g8r8a8Srgb":                         vk.FormatB8g8r8a8Srgb,
		"FormatA8b8g8r8UnormPack32":                  vk.FormatA8b8g8r8UnormPack32,
		"FormatA8b8g8r8SnormPack32":                  vk.FormatA8b8g8r8SnormPack32,
		"FormatA8b8g8r8UscaledPack32":                vk.FormatA8b8g8r8UscaledPack32,
		"FormatA8b8g8r8SscaledPack32":                vk.FormatA8b8g8r8SscaledPack32,
		"FormatA8b8g8r8UintPack32":                   vk.FormatA8b8g8r8UintPack32,
		"FormatA8b8g8r8SintPack32":                   vk.FormatA8b8g8r8SintPack32,
		"FormatA8b8g8r8SrgbPack32":                   vk.FormatA8b8g8r8SrgbPack32,
		"FormatA2r10g10b10UnormPack32":               vk.FormatA2r10g10b10UnormPack32,
		"FormatA2r10g10b10SnormPack32":               vk.FormatA2r10g10b10SnormPack32,
		"FormatA2r10g10b10UscaledPack32":             vk.FormatA2r10g10b10UscaledPack32,
		"FormatA2r10g10b10SscaledPack32":             vk.FormatA2r10g10b10SscaledPack32,
		"FormatA2r10g10b10UintPack32":                vk.FormatA2r10g10b10UintPack32,
		"FormatA2r10g10b10SintPack32":                vk.FormatA2r10g10b10SintPack32,
		"FormatA2b10g10r10UnormPack32":               vk.FormatA2b10g10r10UnormPack32,
		"FormatA2b10g10r10SnormPack32":               vk.FormatA2b10g10r10SnormPack32,
		"FormatA2b10g10r10UscaledPack32":             vk.FormatA2b10g10r10UscaledPack32,
		"FormatA2b10g10r10SscaledPack32":             vk.FormatA2b10g10r10SscaledPack32,
		"FormatA2b10g10r10UintPack32":                vk.FormatA2b10g10r10UintPack32,
		"FormatA2b10g10r10SintPack32":                vk.FormatA2b10g10r10SintPack32,
		"FormatR16Unorm":                             vk.FormatR16Unorm,
		"FormatR16Snorm":                             vk.FormatR16Snorm,
		"FormatR16Uscaled":                           vk.FormatR16Uscaled,
		"FormatR16Sscaled":                           vk.FormatR16Sscaled,
		"FormatR16Uint":                              vk.FormatR16Uint,
		"FormatR16Sint":                              vk.FormatR16Sint,
		"FormatR16Sfloat":                            vk.FormatR16Sfloat,
		"FormatR16g16Unorm":                          vk.FormatR16g16Unorm,
		"FormatR16g16Snorm":                          vk.FormatR16g16Snorm,
		"FormatR16g16Uscaled":                        vk.FormatR16g16Uscaled,
		"FormatR16g16Sscaled":                        vk.FormatR16g16Sscaled,
		"FormatR16g16Uint":                           vk.FormatR16g16Uint,
		"FormatR16g16Sint":                           vk.FormatR16g16Sint,
		"FormatR16g16Sfloat":                         vk.FormatR16g16Sfloat,
		"FormatR16g16b16Unorm":                       vk.FormatR16g16b16Unorm,
		"FormatR16g16b16Snorm":                       vk.FormatR16g16b16Snorm,
		"FormatR16g16b16Uscaled":                     vk.FormatR16g16b16Uscaled,
		"FormatR16g16b16Sscaled":                     vk.FormatR16g16b16Sscaled,
		"FormatR16g16b16Uint":                        vk.FormatR16g16b16Uint,
		"FormatR16g16b16Sint":                        vk.FormatR16g16b16Sint,
		"FormatR16g16b16Sfloat":                      vk.FormatR16g16b16Sfloat,
		"FormatR16g16b16a16Unorm":                    vk.FormatR16g16b16a16Unorm,
		"FormatR16g16b16a16Snorm":                    vk.FormatR16g16b16a16Snorm,
		"FormatR16g16b16a16Uscaled":                  vk.FormatR16g16b16a16Uscaled,
		"FormatR16g16b16a16Sscaled":                  vk.FormatR16g16b16a16Sscaled,
		"FormatR16g16b16a16Uint":                     vk.FormatR16g16b16a16Uint,
		"FormatR16g16b16a16Sint":                     vk.FormatR16g16b16a16Sint,
		"FormatR16g16b16a16Sfloat":                   vk.FormatR16g16b16a16Sfloat,
		"FormatR32Uint":                              vk.FormatR32Uint,
		"FormatR32Sint":                              vk.FormatR32Sint,
		"FormatR32Sfloat":                            vk.FormatR32Sfloat,
		"FormatR32g32Uint":                           vk.FormatR32g32Uint,
		"FormatR32g32Sint":                           vk.FormatR32g32Sint,
		"FormatR32g32Sfloat":                         vk.FormatR32g32Sfloat,
		"FormatR32g32b32Uint":                        vk.FormatR32g32b32Uint,
		"FormatR32g32b32Sint":                        vk.FormatR32g32b32Sint,
		"FormatR32g32b32Sfloat":                      vk.FormatR32g32b32Sfloat,
		"FormatR32g32b32a32Uint":                     vk.FormatR32g32b32a32Uint,
		"FormatR32g32b32a32Sint":                     vk.FormatR32g32b32a32Sint,
		"FormatR32g32b32a32Sfloat":                   vk.FormatR32g32b32a32Sfloat,
		"FormatR64Uint":                              vk.FormatR64Uint,
		"FormatR64Sint":                              vk.FormatR64Sint,
		"FormatR64Sfloat":                            vk.FormatR64Sfloat,
		"FormatR64g64Uint":                           vk.FormatR64g64Uint,
		"FormatR64g64Sint":                           vk.FormatR64g64Sint,
		"FormatR64g64Sfloat":                         vk.FormatR64g64Sfloat,
		"FormatR64g64b64Uint":                        vk.FormatR64g64b64Uint,
		"FormatR64g64b64Sint":                        vk.FormatR64g64b64Sint,
		"FormatR64g64b64Sfloat":                      vk.FormatR64g64b64Sfloat,
		"FormatR64g64b64a64Uint":                     vk.FormatR64g64b64a64Uint,
		"FormatR64g64b64a64Sint":                     vk.FormatR64g64b64a64Sint,
		"FormatR64g64b64a64Sfloat":                   vk.FormatR64g64b64a64Sfloat,
		"FormatB10g11r11UfloatPack32":                vk.FormatB10g11r11UfloatPack32,
		"FormatE5b9g9r9UfloatPack32":                 vk.FormatE5b9g9r9UfloatPack32,
		"FormatD16Unorm":                             vk.FormatD16Unorm,
		"FormatX8D24UnormPack32":                     vk.FormatX8D24UnormPack32,
		"FormatD32Sfloat":                            vk.FormatD32Sfloat,
		"FormatS8Uint":                               vk.FormatS8Uint,
		"FormatD16UnormS8Uint":                       vk.FormatD16UnormS8Uint,
		"FormatD24UnormS8Uint":                       vk.FormatD24UnormS8Uint,
		"FormatD32SfloatS8Uint":                      vk.FormatD32SfloatS8Uint,
		"FormatBc1RgbUnormBlock":                     vk.FormatBc1RgbUnormBlock,
		"FormatBc1RgbSrgbBlock":                      vk.FormatBc1RgbSrgbBlock,
		"FormatBc1RgbaUnormBlock":                    vk.FormatBc1RgbaUnormBlock,
		"FormatBc1RgbaSrgbBlock":                     vk.FormatBc1RgbaSrgbBlock,
		"FormatBc2UnormBlock":                        vk.FormatBc2UnormBlock,
		"FormatBc2SrgbBlock":                         vk.FormatBc2SrgbBlock,
		"FormatBc3UnormBlock":                        vk.FormatBc3UnormBlock,
		"FormatBc3SrgbBlock":                         vk.FormatBc3SrgbBlock,
		"FormatBc4UnormBlock":                        vk.FormatBc4UnormBlock,
		"FormatBc4SnormBlock":                        vk.FormatBc4SnormBlock,
		"FormatBc5UnormBlock":                        vk.FormatBc5UnormBlock,
		"FormatBc5SnormBlock":                        vk.FormatBc5SnormBlock,
		"FormatBc6hUfloatBlock":                      vk.FormatBc6hUfloatBlock,
		"FormatBc6hSfloatBlock":                      vk.FormatBc6hSfloatBlock,
		"FormatBc7UnormBlock":                        vk.FormatBc7UnormBlock,
		"FormatBc7SrgbBlock":                         vk.FormatBc7SrgbBlock,
		"FormatEtc2R8g8b8UnormBlock":                 vk.FormatEtc2R8g8b8UnormBlock,
		"FormatEtc2R8g8b8SrgbBlock":                  vk.FormatEtc2R8g8b8SrgbBlock,
		"FormatEtc2R8g8b8a1UnormBlock":               vk.FormatEtc2R8g8b8a1UnormBlock,
		"FormatEtc2R8g8b8a1SrgbBlock":                vk.FormatEtc2R8g8b8a1SrgbBlock,
		"FormatEtc2R8g8b8a8UnormBlock":               vk.FormatEtc2R8g8b8a8UnormBlock,
		"FormatEtc2R8g8b8a8SrgbBlock":                vk.FormatEtc2R8g8b8a8SrgbBlock,
		"FormatEacR11UnormBlock":                     vk.FormatEacR11UnormBlock,
		"FormatEacR11SnormBlock":                     vk.FormatEacR11SnormBlock,
		"FormatEacR11g11UnormBlock":                  vk.FormatEacR11g11UnormBlock,
		"FormatEacR11g11SnormBlock":                  vk.FormatEacR11g11SnormBlock,
		"FormatAstc4x4UnormBlock":                    vk.FormatAstc4x4UnormBlock,
		"FormatAstc4x4SrgbBlock":                     vk.FormatAstc4x4SrgbBlock,
		"FormatAstc5x4UnormBlock":                    vk.FormatAstc5x4UnormBlock,
		"FormatAstc5x4SrgbBlock":                     vk.FormatAstc5x4SrgbBlock,
		"FormatAstc5x5UnormBlock":                    vk.FormatAstc5x5UnormBlock,
		"FormatAstc5x5SrgbBlock":                     vk.FormatAstc5x5SrgbBlock,
		"FormatAstc6x5UnormBlock":                    vk.FormatAstc6x5UnormBlock,
		"FormatAstc6x5SrgbBlock":                     vk.FormatAstc6x5SrgbBlock,
		"FormatAstc6x6UnormBlock":                    vk.FormatAstc6x6UnormBlock,
		"FormatAstc6x6SrgbBlock":                     vk.FormatAstc6x6SrgbBlock,
		"FormatAstc8x5UnormBlock":                    vk.FormatAstc8x5UnormBlock,
		"FormatAstc8x5SrgbBlock":                     vk.FormatAstc8x5SrgbBlock,
		"FormatAstc8x6UnormBlock":                    vk.FormatAstc8x6UnormBlock,
		"FormatAstc8x6SrgbBlock":                     vk.FormatAstc8x6SrgbBlock,
		"FormatAstc8x8UnormBlock":                    vk.FormatAstc8x8UnormBlock,
		"FormatAstc8x8SrgbBlock":                     vk.FormatAstc8x8SrgbBlock,
		"FormatAstc10x5UnormBlock":                   vk.FormatAstc10x5UnormBlock,
		"FormatAstc10x5SrgbBlock":                    vk.FormatAstc10x5SrgbBlock,
		"FormatAstc10x6UnormBlock":                   vk.FormatAstc10x6UnormBlock,
		"FormatAstc10x6SrgbBlock":                    vk.FormatAstc10x6SrgbBlock,
		"FormatAstc10x8UnormBlock":                   vk.FormatAstc10x8UnormBlock,
		"FormatAstc10x8SrgbBlock":                    vk.FormatAstc10x8SrgbBlock,
		"FormatAstc10x10UnormBlock":                  vk.FormatAstc10x10UnormBlock,
		"FormatAstc10x10SrgbBlock":                   vk.FormatAstc10x10SrgbBlock,
		"FormatAstc12x10UnormBlock":                  vk.FormatAstc12x10UnormBlock,
		"FormatAstc12x10SrgbBlock":                   vk.FormatAstc12x10SrgbBlock,
		"FormatAstc12x12UnormBlock":                  vk.FormatAstc12x12UnormBlock,
		"FormatAstc12x12SrgbBlock":                   vk.FormatAstc12x12SrgbBlock,
		"FormatG8b8g8r8422Unorm":                     vk.FormatG8b8g8r8422Unorm,
		"FormatB8g8r8g8422Unorm":                     vk.FormatB8g8r8g8422Unorm,
		"FormatG8B8R83plane420Unorm":                 vk.FormatG8B8R83plane420Unorm,
		"FormatG8B8r82plane420Unorm":                 vk.FormatG8B8r82plane420Unorm,
		"FormatG8B8R83plane422Unorm":                 vk.FormatG8B8R83plane422Unorm,
		"FormatG8B8r82plane422Unorm":                 vk.FormatG8B8r82plane422Unorm,
		"FormatG8B8R83plane444Unorm":                 vk.FormatG8B8R83plane444Unorm,
		"FormatR10x6UnormPack16":                     vk.FormatR10x6UnormPack16,
		"FormatR10x6g10x6Unorm2pack16":               vk.FormatR10x6g10x6Unorm2pack16,
		"FormatR10x6g10x6b10x6a10x6Unorm4pack16":     vk.FormatR10x6g10x6b10x6a10x6Unorm4pack16,
		"FormatG10x6b10x6g10x6r10x6422Unorm4pack16":  vk.FormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		"FormatB10x6g10x6r10x6g10x6422Unorm4pack16":  vk.FormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		"FormatG10x6B10x6R10x63plane420Unorm3pack16": vk.FormatG10x6B10x6R10x63plane420Unorm3pack16,
		"FormatG10x6B10x6r10x62plane420Unorm3pack16": vk.FormatG10x6B10x6r10x62plane420Unorm3pack16,
		"FormatG10x6B10x6R10x63plane422Unorm3pack16": vk.FormatG10x6B10x6R10x63plane422Unorm3pack16,
		"FormatG10x6B10x6r10x62plane422Unorm3pack16": vk.FormatG10x6B10x6r10x62plane422Unorm3pack16,
		"FormatG10x6B10x6R10x63plane444Unorm3pack16": vk.FormatG10x6B10x6R10x63plane444Unorm3pack16,
		"FormatR12x4UnormPack16":                     vk.FormatR12x4UnormPack16,
		//"FormatR12x4g12x4Unorm2pack16":               vk.FormatR12x4g12x4Unorm2pack16,
		"FormatR12x4g12x4b12x4a12x4Unorm4pack16":     vk.FormatR12x4g12x4b12x4a12x4Unorm4pack16,
		"FormatG12x4b12x4g12x4r12x4422Unorm4pack16":  vk.FormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		"FormatB12x4g12x4r12x4g12x4422Unorm4pack16":  vk.FormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		"FormatG12x4B12x4R12x43plane420Unorm3pack16": vk.FormatG12x4B12x4R12x43plane420Unorm3pack16,
		"FormatG12x4B12x4r12x42plane420Unorm3pack16": vk.FormatG12x4B12x4r12x42plane420Unorm3pack16,
		"FormatG12x4B12x4R12x43plane422Unorm3pack16": vk.FormatG12x4B12x4R12x43plane422Unorm3pack16,
		"FormatG12x4B12x4r12x42plane422Unorm3pack16": vk.FormatG12x4B12x4r12x42plane422Unorm3pack16,
		"FormatG12x4B12x4R12x43plane444Unorm3pack16": vk.FormatG12x4B12x4R12x43plane444Unorm3pack16,
		"FormatG16b16g16r16422Unorm":                 vk.FormatG16b16g16r16422Unorm,
		"FormatB16g16r16g16422Unorm":                 vk.FormatB16g16r16g16422Unorm,
		"FormatG16B16R163plane420Unorm":              vk.FormatG16B16R163plane420Unorm,
		"FormatG16B16r162plane420Unorm":              vk.FormatG16B16r162plane420Unorm,
		"FormatG16B16R163plane422Unorm":              vk.FormatG16B16R163plane422Unorm,
		"FormatG16B16r162plane422Unorm":              vk.FormatG16B16r162plane422Unorm,
		"FormatG16B16R163plane444Unorm":              vk.FormatG16B16R163plane444Unorm,
		"FormatPvrtc12bppUnormBlockImg":              vk.FormatPvrtc12bppUnormBlockImg,
		"FormatPvrtc14bppUnormBlockImg":              vk.FormatPvrtc14bppUnormBlockImg,
		"FormatPvrtc22bppUnormBlockImg":              vk.FormatPvrtc22bppUnormBlockImg,
		"FormatPvrtc24bppUnormBlockImg":              vk.FormatPvrtc24bppUnormBlockImg,
		"FormatPvrtc12bppSrgbBlockImg":               vk.FormatPvrtc12bppSrgbBlockImg,
		"FormatPvrtc14bppSrgbBlockImg":               vk.FormatPvrtc14bppSrgbBlockImg,
		//"FormatPvrtc22bppSrgbBlockImg":               vk.FormatPvrtc22bppSrgbBlockImg,
		"FormatPvrtc24bppSrgbBlockImg": vk.FormatPvrtc24bppSrgbBlockImg,
		"FormatBeginRange":             vk.FormatBeginRange,
		"FormatEndRange":               vk.FormatEndRange,
		//"FormatRangeSize":              vk.FormatRangeSize,
	}

	var formatPropertiesFlags = map[string]vk.FormatFeatureFlagBits{
		"FormatFeatureSampledImageBit":                                                     vk.FormatFeatureSampledImageBit,
		"FormatFeatureStorageImageBit":                                                     vk.FormatFeatureStorageImageBit,
		"FormatFeatureStorageImageAtomicBit":                                               vk.FormatFeatureStorageImageAtomicBit,
		"FormatFeatureUniformTexelBufferBit":                                               vk.FormatFeatureUniformTexelBufferBit,
		"FormatFeatureStorageTexelBufferBit":                                               vk.FormatFeatureStorageTexelBufferBit,
		"FormatFeatureStorageTexelBufferAtomicBit":                                         vk.FormatFeatureStorageTexelBufferAtomicBit,
		"FormatFeatureVertexBufferBit":                                                     vk.FormatFeatureVertexBufferBit,
		"FormatFeatureColorAttachmentBit":                                                  vk.FormatFeatureColorAttachmentBit,
		"FormatFeatureColorAttachmentBlendBit":                                             vk.FormatFeatureColorAttachmentBlendBit,
		"FormatFeatureDepthStencilAttachmentBit":                                           vk.FormatFeatureDepthStencilAttachmentBit,
		"FormatFeatureBlitSrcBit":                                                          vk.FormatFeatureBlitSrcBit,
		"FormatFeatureBlitDstBit":                                                          vk.FormatFeatureBlitDstBit,
		"FormatFeatureSampledImageFilterLinearBit":                                         vk.FormatFeatureSampledImageFilterLinearBit,
		"FormatFeatureTransferSrcBit":                                                      vk.FormatFeatureTransferSrcBit,
		"FormatFeatureTransferDstBit":                                                      vk.FormatFeatureTransferDstBit,
		"FormatFeatureMidpointChromaSamplesBit":                                            vk.FormatFeatureMidpointChromaSamplesBit,
		"FormatFeatureSampledImageYcbcrConversionLinearFilterBit":                          vk.FormatFeatureSampledImageYcbcrConversionLinearFilterBit,
		"FormatFeatureSampledImageYcbcrConversionSeparateReconstructionFilterBit":          vk.FormatFeatureSampledImageYcbcrConversionSeparateReconstructionFilterBit,
		"FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitBit":          vk.FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitBit,
		"FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitForceableBit": vk.FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitForceableBit,
		"FormatFeatureDisjointBit":                                                         vk.FormatFeatureDisjointBit,
		"FormatFeatureCositedChromaSamplesBit":                                             vk.FormatFeatureCositedChromaSamplesBit,
		"FormatFeatureSampledImageFilterCubicBitImg":                                       vk.FormatFeatureSampledImageFilterCubicBitImg,
		"FormatFeatureSampledImageFilterMinmaxBit":                                         vk.FormatFeatureSampledImageFilterMinmaxBit,
	}
	// For 2D image with TillingLinear
	fmt.Println("\n\nChecking supported Iamge/Texture Formats for 2D image with TillingLinear ......")
	var imageType vk.ImageType = vk.ImageType2d
	var imageTiling vk.ImageTiling = vk.ImageTilingLinear
	//var imageUsageFlags vk.ImageUsageFlagBits
	var imageCreateFlags vk.ImageCreateFlags
	var imageFormatProperties vk.ImageFormatProperties
	for key, format := range formats {
		result := vk.GetPhysicalDeviceImageFormatProperties(physicalDevice, format, imageType, imageTiling, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit), imageCreateFlags, &imageFormatProperties)
		if result == vk.Success {
			fmt.Printf("Supported Format [%v]: FlagValue [%v]\nFollowing are the Features/Properties available in GPU against above format\n", key, format)
			fmt.Printf("\t*\tImageType: %v\t\t \t\t\n", imageType)
			fmt.Printf("\t*\tImageTiling: %v\t\t \t\t\n", imageTiling)
			imageFormatProperties.Deref()
			extent := imageFormatProperties.MaxExtent
			maxmipmaplevel := imageFormatProperties.MaxMipLevels
			maxarraylayers := imageFormatProperties.MaxArrayLayers
			maxResourceSize := imageFormatProperties.MaxResourceSize
			extent.Deref()
			fmt.Printf("\t*\tExtent: \t\t \t\t\n")
			fmt.Printf("\t\t\tWidth: %v\t\t\n", extent.Width)
			fmt.Printf("\t\t\tHeight: %v\t\t\n", extent.Height)
			fmt.Printf("\t\t\tDepth: %v\t\t\n", extent.Depth)
			fmt.Printf("\t*\tMaxMipMap Levels: %v\t\t \t\t\n", maxmipmaplevel)
			fmt.Printf("\t*\tMaxArrayLayers: %v\t\t \t\t\n", maxarraylayers)
			fmt.Printf("\t*\tMaxResourceSize: %v\t\t \t\t\n", maxResourceSize)
			// Printing FormatProperties
			fmt.Printf("\t*\tFormatProperties: \t\t \t\t\n")
			var formatProperties vk.FormatProperties
			vk.GetPhysicalDeviceFormatProperties(physicalDevice, format, &formatProperties)
			formatProperties.Deref()
			fmt.Printf("\t\t\t*\tAll LinearTilingFeatures present: %v\t\t \t\t\n", vk.FormatFeatureFlagBits(formatProperties.LinearTilingFeatures))
			for key, flag := range formatPropertiesFlags {
				if flag&vk.FormatFeatureFlagBits(formatProperties.LinearTilingFeatures) != 0x00000000 {
					fmt.Printf("\t\t\t\t\t*\tLinearTilingFeatures: vk.%v[%v]\t\t \t\t\n", key, flag&vk.FormatFeatureFlagBits(formatProperties.LinearTilingFeatures))
				}
			}
			fmt.Printf("\t\t\t*\tAll OptimalTilingFeatures present: %v\t\t \t\t\n", vk.FormatFeatureFlagBits(formatProperties.OptimalTilingFeatures))
			for key, flag := range formatPropertiesFlags {
				if flag&vk.FormatFeatureFlagBits(formatProperties.OptimalTilingFeatures) != 0x00000000 {
					fmt.Printf("\t\t\t\t\t*\tOptimalTilingFeatures: vk.%v[%v]\t\t \t\t\n", key, flag&vk.FormatFeatureFlagBits(formatProperties.OptimalTilingFeatures))
				}
			}
			fmt.Printf("\t\t\t*\tAll BufferFeatures present: %v\t\t \t\t\n", vk.FormatFeatureFlagBits(formatProperties.BufferFeatures))
			for key, flag := range formatPropertiesFlags {
				if flag&vk.FormatFeatureFlagBits(formatProperties.BufferFeatures) != 0x00000000 {
					fmt.Printf("\t\t\t\t\t*\tBufferFeatures: vk.%v[%v]\t\t \t\t\n", key, flag&vk.FormatFeatureFlagBits(formatProperties.BufferFeatures))
				}
			}
		}

	}
	// For 3D image with TillingLinear
	fmt.Println("\n\nChecking supported Image/Texture Formats for 3D image with TillingLinear ......")
	imageType = vk.ImageType3d
	imageTiling = vk.ImageTilingLinear
	//var imageUsageFlags vk.ImageUsageFlagBits
	//var imageCreateFlags vk.ImageCreateFlags
	//var imageFormatProperties vk.ImageFormatProperties
	for key, format := range formats {
		result := vk.GetPhysicalDeviceImageFormatProperties(physicalDevice, format, imageType, imageTiling, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit), imageCreateFlags, &imageFormatProperties)
		if result == vk.Success {
			fmt.Printf("Supported Format [%v]: FlagValue [%v]\nFollowing are the Features/Properties available in GPU against above format\n", key, format)
			fmt.Printf("\t*\tImageType: %v\t\t \t\t\n", imageType)
			fmt.Printf("\t*\tImageTiling: %v\t\t \t\t\n", imageTiling)
			imageFormatProperties.Deref()
			extent := imageFormatProperties.MaxExtent
			maxmipmaplevel := imageFormatProperties.MaxMipLevels
			maxarraylayers := imageFormatProperties.MaxArrayLayers
			maxResourceSize := imageFormatProperties.MaxResourceSize
			extent.Deref()
			fmt.Printf("\t*\tExtent: \t\t \t\t\n")
			fmt.Printf("\t\t\tWidth: %v\t\t\n", extent.Width)
			fmt.Printf("\t\t\tHeight: %v\t\t\n", extent.Height)
			fmt.Printf("\t\t\tDepth: %v\t\t\n", extent.Depth)
			fmt.Printf("\t*\tMaxMipMap Levels: %v\t\t \t\t\n", maxmipmaplevel)
			fmt.Printf("\t*\tMaxArrayLayers: %v\t\t \t\t\n", maxarraylayers)
			fmt.Printf("\t*\tMaxResourceSize: %v\t\t \t\t\n", maxResourceSize)
			// Printing FormatProperties
			fmt.Printf("\t*\tFormatProperties: \t\t \t\t\n")
			var formatProperties vk.FormatProperties
			vk.GetPhysicalDeviceFormatProperties(physicalDevice, format, &formatProperties)
			formatProperties.Deref()
			fmt.Printf("\t\t\t*\tAll LinearTilingFeatures present: %v\t\t \t\t\n", vk.FormatFeatureFlagBits(formatProperties.LinearTilingFeatures))
			for key, flag := range formatPropertiesFlags {
				if flag&vk.FormatFeatureFlagBits(formatProperties.LinearTilingFeatures) != 0x00000000 {
					fmt.Printf("\t\t\t\t\t*\tLinearTilingFeatures: vk.%v[%v]\t\t \t\t\n", key, flag&vk.FormatFeatureFlagBits(formatProperties.LinearTilingFeatures))
				}
			}
			fmt.Printf("\t\t\t*\tAll OptimalTilingFeatures present: %v\t\t \t\t\n", vk.FormatFeatureFlagBits(formatProperties.OptimalTilingFeatures))
			for key, flag := range formatPropertiesFlags {
				if flag&vk.FormatFeatureFlagBits(formatProperties.OptimalTilingFeatures) != 0x00000000 {
					fmt.Printf("\t\t\t\t\t*\tOptimalTilingFeatures: vk.%v[%v]\t\t \t\t\n", key, flag&vk.FormatFeatureFlagBits(formatProperties.OptimalTilingFeatures))
				}
			}
			fmt.Printf("\t\t\t*\tAll BufferFeatures present: %v\t\t \t\t\n", vk.FormatFeatureFlagBits(formatProperties.BufferFeatures))
			for key, flag := range formatPropertiesFlags {
				if flag&vk.FormatFeatureFlagBits(formatProperties.BufferFeatures) != 0x00000000 {
					fmt.Printf("\t\t\t\t\t*\tBufferFeatures: vk.%v[%v]\t\t \t\t\n", key, flag&vk.FormatFeatureFlagBits(formatProperties.BufferFeatures))
				}
			}
		}
	}
}
