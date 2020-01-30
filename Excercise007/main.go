package main

import (
	"fmt"

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
	var srcBuffer, dstBuffer *vk.Buffer
	var imageFormatProperties vk.ImageFormatProperties
	var imageBuffer *vk.Image

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

	physicalDevices = getPhysicalDevices(instance)
	physicalDeviceProperties = getPhysicalDeviceProperties(physicalDevices[0])
	physicalDeviceFeatures = getPhysicalDeviceFeatures(physicalDevices[0])
	memoryProperties = getPhysicalDeviceMemoryProperties(physicalDevices[0])
	pQueueFamilyProperties = getPhysicalDeviceQueueFamilyProperties(physicalDevices[0])
	pLogicalDevice = createDevice(physicalDevices[0], physicalDeviceFeatures)
	getInstanceLayerProperties()
	getDeviceLayerProperties(physicalDevices[0])
	getInstanceExtensionProperties()
	getDeviceExtensionProperties(physicalDevices[0])
	deviceWaitTillComplete(pLogicalDevice)
	commandPool = createCommandPool(*pLogicalDevice)
	commandBuffers = allocateCommandBuffers(*pLogicalDevice, *commandPool, 1)
	srcBuffer = createBuffer(*pLogicalDevice)
	imageFormatProperties = getPhysicalDeviceImageProperties(physicalDevices[0])
	imageBuffer = createImageBuffer(pLogicalDevice)
	checkSupportedImageFormat(physicalDevices[0])
	// Get the memory properties of the physical device.
	vk.GetPhysicalDeviceMemoryProperties(physicalDevices[0], &memoryProperties)

	// Verbose - Please don't remove, igrnoe
	physicalDeviceProperties.Deref()
	fmt.Println(physicalDevices)
	fmt.Println(vk.ToString(physicalDeviceProperties.DeviceName[:]))
	fmt.Println(physicalDeviceFeatures)
	fmt.Println(memoryProperties)
	fmt.Println(pQueueFamilyProperties)
	uitlPrintDeviceQueueFamilyProperties(pQueueFamilyProperties)
	fmt.Printf("%T, %v", pLogicalDevice, pLogicalDevice)
	fmt.Println(srcBuffer, dstBuffer)
	fmt.Println(&imageFormatProperties)
	fmt.Println(commandPool)
	fmt.Println(commandBuffers)
	fmt.Println(imageBuffer)

	//Cleaningup code
	vk.FreeCommandBuffers(*pLogicalDevice, *commandPool, 1, commandBuffers)
	vk.DestroyCommandPool(*pLogicalDevice, *commandPool, nil)
}

func recordCommandIntoCommandBuffer(commandBuffer vk.CommandBuffer) {
	var commandBufferBeginInfo = vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: 0x0,
	}
	result := vk.BeginCommandBuffer(commandBuffer, &commandBufferBeginInfo)
	if result != vk.Success {
		fmt.Printf("Failed to begin command buffer with error : %v", result)
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
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: 0,
	}
	result := vk.CreateCommandPool(pLogicalDevice, &commandPollCreateInfo, nil, &commandPool)
	if result != vk.Success {
		fmt.Printf("Failed to create command pool with error : %v", result)
	}
	return &commandPool
}

func createImageBuffer(pLogicalDevice *vk.Device) *vk.Image {
	fmt.Println("Creating image buffer...........")
	var imageBuffer vk.Image
	var extent = vk.Extent3D{
		Width:  1024,
		Height: 1024,
		Depth:  1,
	}
	var imageCreateInfo = vk.ImageCreateInfo{
		SType:                 vk.StructureTypeImageCreateInfo,
		ImageType:             vk.ImageType2d,
		Format:                vk.FormatB8g8r8a8Srgb,
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

func getPhysicalDeviceImageProperties(physicalDevice vk.PhysicalDevice) vk.ImageFormatProperties {
	fmt.Println("Physical device supported image format for 3D data of sRGB FormatB8g8r8Unorm Format........") // Another popular format vk.FormatR8g8b8Unorm
	var imageType vk.ImageType = vk.ImageType3d
	var imageTiling vk.ImageTiling = vk.ImageTilingLinear
	//var imageUsageFlags vk.ImageUsageFlagBits
	var imageCreateFlags vk.ImageCreateFlags
	var imageFormatProperties vk.ImageFormatProperties
	result := vk.GetPhysicalDeviceImageFormatProperties(physicalDevice, vk.FormatB8g8r8Unorm, imageType, imageTiling, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit), imageCreateFlags, &imageFormatProperties)
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
		Size:        1024, // 1Kb
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
		QueueCount: 1,
		//QueueFamilyIndex: 0,
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

func uitlPrintDeviceQueueFamilyProperties(qfs []vk.QueueFamilyProperties) {
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
	var imageType vk.ImageType = vk.ImageType3d
	var imageTiling vk.ImageTiling = vk.ImageTilingLinear
	//var imageUsageFlags vk.ImageUsageFlagBits
	var imageCreateFlags vk.ImageCreateFlags
	var imageFormatProperties vk.ImageFormatProperties
	for key, format := range formats {
		fmt.Printf("Supported Format [%v]: %v\n", key, format)
		result := vk.GetPhysicalDeviceImageFormatProperties(physicalDevice, format, imageType, imageTiling, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit), imageCreateFlags, &imageFormatProperties)
		if result == vk.Success {
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
		}
	}
}
