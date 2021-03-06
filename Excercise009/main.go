package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

func main() {
	glfw.Init()
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	vk.Init()

	//1. Instance
	//	1. Create Application Struct
	//	2. Create Instance
	//	3. Create Window
	//	4. Create Surface
	var instance vk.Instance
	var appInfo vk.ApplicationInfo = vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.ApiVersion10,
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	var layers = []string{"VK_LAYER_NV_optimus\x00"}
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PNext:                   nil,
		Flags:                   0,
		PApplicationInfo:        &appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	result := vk.CreateInstance(&instanceInfo, nil, &instance)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating instance: %v", result))
	}
	var window *glfw.Window
	var err error
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err = glfw.CreateWindow(640, 480, "Vulkan Info", nil, nil)
	if err != nil {
		fmt.Println("Failed to create window with error :", err)
	}
	var surface vk.Surface
	pSurface, glfwError := window.CreateWindowSurface(instance, nil)
	if glfwError != nil {
		fmt.Println("Failed to create window surface with error :", glfwError)
	}
	surface = vk.SurfaceFromPointer(pSurface)
	fmt.Println(surface)
	//2. Logical Device
	//	1. Get All Physical Devices GPU
	//  2. List Physical Devices GPU properties
	//  3. Get MemoryProprties of Selected Physical Device
	//	4. Get Device Extension Properties
	//	5. Create Logical Device
	var deviceCount uint32
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	fmt.Printf("\tFound total %v Physical Device(s).....\n", deviceCount)
	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	var physicalDevice = physicalDevices[0]
	var physicalDeviceMemoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(physicalDevice, &physicalDeviceMemoryProperties)
	deviceQueueCreateInfo := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: 0,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	var deviceProperties vk.PhysicalDeviceProperties
	vk.GetPhysicalDeviceProperties(physicalDevice, &deviceProperties)
	deviceProperties.Deref()
	fmt.Println("Driver Details...............")
	fmt.Println("\t* Driver:\t", deviceProperties.DriverVersion)
	fmt.Println("\t* Type:\t\t", deviceProperties.DeviceType)
	fmt.Printf("\t* Version: \t%v.%v.%v\n", (deviceProperties.ApiVersion>>22)&0x3FF, (deviceProperties.ApiVersion>>22)&0x3FF, deviceProperties.ApiVersion&0xFFF)
	fmt.Println("\t* Name:\t", string(deviceProperties.DeviceName[:]))
	var deviceExtensionCount uint32
	result = vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &deviceExtensionCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	var deviceExtensions = make([]vk.ExtensionProperties, deviceExtensionCount)
	var deviceExtensionNames = make([]string, deviceExtensionCount)
	for idx, extension := range deviceExtensions {
		extension.Deref()
		deviceExtensionNames[idx] = fmt.Sprintf("%v\x00", extension.ExtensionName[:])
	}
	result = vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &deviceExtensionCount, deviceExtensions)
	var deviceFeatures = make([]vk.PhysicalDeviceFeatures, 1)
	deviceFeatures[0].ShaderClipDistance = vk.True
	var deviceCreateInfo vk.DeviceCreateInfo = vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(deviceQueueCreateInfo)),
		PQueueCreateInfos:       deviceQueueCreateInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(deviceExtensionNames)),
		PpEnabledExtensionNames: deviceExtensionNames,
		PEnabledFeatures:        deviceFeatures,
	}
	var logicalDevice vk.Device
	result = vk.CreateDevice(physicalDevice, &deviceCreateInfo, nil, &logicalDevice)

	//3. Create SwapChain
	//	1. Get surface capabilities
	//  2. Create Swapchain
	var physicalDeviceSurfaceCapabilities vk.SurfaceCapabilities
	result = vk.GetPhysicalDeviceSurfaceCapabilities(physicalDevice, surface, &physicalDeviceSurfaceCapabilities)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device surface capabilities: %v", result))
		panic(result)
	}
	var formatCount uint32
	vk.GetPhysicalDeviceSurfaceFormats(physicalDevice, surface, &formatCount, nil)
	formats := make([]vk.SurfaceFormat, formatCount)
	vk.GetPhysicalDeviceSurfaceFormats(physicalDevice, surface, &formatCount, formats)

	chosenFormat := -1
	for i := 0; i < int(formatCount); i++ {
		formats[i].Deref()
		log.Printf("physical device surface formats and colorspace available %v : % v\n", formats[0].Format, formats[1].Format)
		if formats[i].Format == vk.FormatB8g8r8a8Unorm ||
			formats[i].Format == vk.FormatR8g8b8a8Unorm {
			chosenFormat = i
			break
		}
	}
	if chosenFormat < 0 {
		fmt.Println("vk.GetPhysicalDeviceSurfaceFormats not found suitable format")
	}
	physicalDeviceSurfaceCapabilities.Deref()
	var surfaceResuloution vk.Extent2D
	surfaceResuloution = physicalDeviceSurfaceCapabilities.CurrentExtent
	surfaceResuloution.Deref()
	formats[chosenFormat].Deref()
	width := surfaceResuloution.Width
	height := surfaceResuloution.Height
	var swapChainInfo = vk.SwapchainCreateInfo{
		SType:                 vk.StructureTypeSwapchainCreateInfo,
		Surface:               surface,
		MinImageCount:         2,
		ImageFormat:           vk.FormatB8g8r8a8Unorm,     //,
		ImageColorSpace:       vk.ColorspaceSrgbNonlinear, //,
		ImageExtent:           surfaceResuloution,
		ImageArrayLayers:      1,
		ImageUsage:            vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		ImageSharingMode:      vk.SharingModeExclusive,
		QueueFamilyIndexCount: 1,
		PQueueFamilyIndices:   []uint32{0},
		PreTransform:          vk.SurfaceTransformIdentityBit,
		CompositeAlpha:        vk.CompositeAlphaOpaqueBit,
		PresentMode:           vk.PresentModeMailbox, //vk.PresentModeFifo, //
		Clipped:               vk.True,
		OldSwapchain:          vk.NullSwapchain,
	}
	var swapChain = make([]vk.Swapchain, 1)
	result = vk.CreateSwapchain(logicalDevice, &swapChainInfo, nil, &swapChain[0])
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating swapchain: %v", result))
		panic(result)
	}
	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	//4. Create Image and ImageView
	//	1. Get coount of images required by swapchain
	//  2. Create Swapchain
	var imageCount uint32
	result = vk.GetSwapchainImages(logicalDevice, swapChain[0], &imageCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting image count from swapchain: %v", result))
		panic(result)
	}
	fmt.Println("Querying Number of Images required by swapchain....", imageCount)
	var images []vk.Image
	images = make([]vk.Image, imageCount)
	result = vk.GetSwapchainImages(logicalDevice, swapChain[0], &imageCount, images)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting images for swapchain: %v", result))
		panic(result)
	}
	var imageViews = make([]vk.ImageView, 2)
	for idx := range imageViews {
		var imageViewCreateInfo = vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			ViewType: vk.ImageViewType2d,
			Format:   vk.FormatB8g8r8a8Unorm,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleR,
				G: vk.ComponentSwizzleG,
				B: vk.ComponentSwizzleB,
				A: vk.ComponentSwizzleA,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			Image: images[idx],
		}
		result = vk.CreateImageView(logicalDevice, &imageViewCreateInfo, nil, &imageViews[idx])
		if result != vk.Success {
			fmt.Println(fmt.Errorf("Failed to create image view : %v", result))
			panic(result)
		}
	}
	//5. Create and Allocate CommandBuffer
	//	1. Get first Queue
	//  2. Create commmand pool
	//  3. Allocate commmand buffer

	var queue vk.Queue
	vk.GetDeviceQueue(logicalDevice, 0, 0, &queue)
	var commandPool vk.CommandPool
	var commandPoolCreateInfo = vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeQueryPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: 0,
	}
	result = vk.CreateCommandPool(logicalDevice, &commandPoolCreateInfo, nil, &commandPool)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Failed to create command pool : %v", result))
		panic(result)
	}
	var commandBuffer []vk.CommandBuffer
	commandBuffer = make([]vk.CommandBuffer, 1)
	var commandBufferAllocateInfo = vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}
	result = vk.AllocateCommandBuffers(logicalDevice, &commandBufferAllocateInfo, commandBuffer)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Failed to allocate command buffer : %v", result))
		panic(result)
	}
	//6. Create FrameBuffer
	//	1. Create Attachment Description and Attachment Reference
	//  2. Create Subpass
	//  3. Create Render Pass
	//  4. Create FrameBuffer
	var attachmentDescriptions []vk.AttachmentDescription
	attachmentDescriptions = make([]vk.AttachmentDescription, 1)
	attachmentDescriptions[0] = vk.AttachmentDescription{
		Format:         vk.FormatB8g8r8a8Unorm,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutColorAttachmentOptimal,
		FinalLayout:    vk.ImageLayoutColorAttachmentOptimal,
	}
	var attachmentReference = make([]vk.AttachmentReference, 1)
	attachmentReference[0].Attachment = 0
	attachmentReference[0].Layout = vk.ImageLayoutColorAttachmentOptimal
	var subpass = make([]vk.SubpassDescription, 1)
	subpass[0] = vk.SubpassDescription{
		PipelineBindPoint:       vk.PipelineBindPointGraphics,
		ColorAttachmentCount:    1,
		PColorAttachments:       attachmentReference,
		PDepthStencilAttachment: nil,
	}
	var renderPassCreateInfo = vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    attachmentDescriptions,
		SubpassCount:    1,
		PSubpasses:      subpass,
	}
	var renderPass vk.RenderPass
	result = vk.CreateRenderPass(logicalDevice, &renderPassCreateInfo, nil, &renderPass)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Failed to create render pass : %v", result))
		panic(result)
	}
	var frameBufferAttachments = make([]vk.ImageView, 1)
	var frameBufferCreateInfo = vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		PNext:           nil,
		RenderPass:      renderPass,
		AttachmentCount: 1,
		PAttachments:    frameBufferAttachments,
		Width:           width,
		Height:          height,
		Layers:          1,
	}
	var frameBuffers = make([]vk.Framebuffer, 2)
	for idx := range frameBuffers {
		result = vk.CreateFramebuffer(logicalDevice, &frameBufferCreateInfo, nil, &frameBuffers[idx])
		if result != vk.Success {
			fmt.Println(fmt.Errorf("Failed to create frame buffers : %v", result))
			panic(result)
		}
	}
	for {
		//7. Display Output
		//	1. Create AcquireNextImage
		//  2. Command Buffer Begin info
		//  3. Create Render Pass
		//  4. Create FrameBuffer
		var nextImageIdx uint32
		result = vk.AcquireNextImage(logicalDevice, swapChain[0], vk.MaxUint64, vk.NullSemaphore, vk.NullFence, &nextImageIdx)
		if result != vk.Success {
			fmt.Println(fmt.Errorf("Failed to acquire next image : %v", result))
			panic(result)
		}
		var commandBufferBeginInfo = vk.CommandBufferBeginInfo{
			SType: vk.StructureTypeCommandBufferBeginInfo,
			Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
		}
		vk.BeginCommandBuffer(commandBuffer[0], &commandBufferBeginInfo)
		var clearValue = []vk.ClearValue{
			vk.NewClearValue([]float32{1.0, 0.0, 0.0, 1.0}),
			vk.NewClearValue([]float32{1.0, 0.0}),
		}
		var RenderPassBeginInfo = vk.RenderPassBeginInfo{
			SType:       vk.StructureTypeRenderPassBeginInfo,
			RenderPass:  renderPass,
			Framebuffer: frameBuffers[nextImageIdx],
		}

		var start = vk.Offset2D{
			X: 0, Y: 0,
		}
		var dim = vk.Extent2D{
			Width:  width,
			Height: height,
		}
		var rect = vk.Rect2D{
			Offset: start,
			Extent: dim,
		}
		RenderPassBeginInfo.RenderArea = rect
		RenderPassBeginInfo.ClearValueCount = 2
		RenderPassBeginInfo.PClearValues = clearValue
		vk.CmdBeginRenderPass(commandBuffer[0], &RenderPassBeginInfo, vk.SubpassContentsBeginRange)
		vk.CmdEndRenderPass(commandBuffer[0])
		vk.EndCommandBuffer(commandBuffer[0])

		var renderFence = make([]vk.Fence, 1)
		var fenceCreateInfo = vk.FenceCreateInfo{
			SType: vk.StructureTypeFenceCreateInfo,
		}
		vk.CreateFence(logicalDevice, &fenceCreateInfo, nil, &renderFence[0])
		var submitInfo = make([]vk.SubmitInfo, 1)
		submitInfo[0] = vk.SubmitInfo{
			SType:                vk.StructureTypeSubmitInfo,
			WaitSemaphoreCount:   0,
			PWaitSemaphores:      nil,
			PWaitDstStageMask:    nil,
			CommandBufferCount:   1,
			PCommandBuffers:      commandBuffer,
			SignalSemaphoreCount: 0,
			PSignalSemaphores:    nil,
		}
		vk.QueueSubmit(queue, 1, submitInfo, renderFence[0])
		const timeoutNano = 10 * 1000 * 1000 * 1000
		vk.WaitForFences(logicalDevice, 1, renderFence, vk.True, timeoutNano)
		vk.DestroyFence(logicalDevice, renderFence[0], nil)
		var presentInfo = vk.PresentInfo{
			SType:              vk.StructureTypePresentInfo,
			WaitSemaphoreCount: 0,
			PWaitSemaphores:    nil,
			PSwapchains:        swapChain,
			PImageIndices:      []uint32{nextImageIdx},
			PResults:           nil,
		}
		vk.QueuePresent(queue, &presentInfo)
		if window.ShouldClose() {
			os.Exit(0)
		}
		glfw.PollEvents()
	}
	//Cleanup
	for idx := range frameBuffers {
		vk.DestroyFramebuffer(logicalDevice, frameBuffers[idx], nil)
	}
	vk.DestroyRenderPass(logicalDevice, renderPass, nil)
	vk.DestroyCommandPool(logicalDevice, commandPool, nil)
	for _, imageView := range imageViews {
		vk.DestroyImageView(logicalDevice, imageView, nil)
	}
	for _, image := range images {
		vk.DestroyImage(logicalDevice, image, nil)
	}
	vk.DestroySwapchain(logicalDevice, swapChain[0], nil)
	vk.DestroySurface(instance, surface, nil)
	vk.DestroyInstance(instance, nil)
	window.Destroy()
	glfw.Terminate()
}
